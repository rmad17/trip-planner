# Quick Start: AI Trip Planner Feature

## ⚡ 5-Minute Setup

### 1. Get Your Anthropic API Key
1. Go to https://console.anthropic.com/
2. Sign up or log in
3. Navigate to API Keys section
4. Create a new API key
5. Copy the key (starts with `sk-ant-`)

### 2. Configure Backend
```bash
cd /home/sourav/projects/trip-planner

# Add to .env file (create if it doesn't exist)
echo "ANTHROPIC_API_KEY=your_api_key_here" >> .env

# Install dependencies (if needed)
go mod tidy

# Start the backend
go run app.go
```

Backend should start on `http://localhost:8080`

### 3. Start Frontend
```bash
cd /home/sourav/projects/trip-planner-fe

# Install dependencies (if needed)
npm install

# Start development server
npm start
```

Frontend should open at `http://localhost:3000`

### 4. Test the Feature

1. **Login** to your account at http://localhost:3000
2. **Click** the "AI Trip Planner" button (orange/coral colored with sparkles icon)
3. **Fill in the form**:
   - Starting From: "San Francisco"
   - Destination: "Tokyo"
   - Start Date: Any future date
   - End Date: 7-14 days after start
   - Number of Travelers: 2
   - Preferences: Select "Culture" and "Food"
   - Pace: "Moderate"

4. **Click** "Generate Trip with AI"
5. **Wait** 10-30 seconds for Claude to generate the trip
6. **Review** the preview with route, budget, and itinerary
7. **Click** "Confirm & Save Trip"
8. **You're done!** You'll be redirected to the full trip details

## 🎯 What You Can Do

### Single Destination Trip
- Source: New York
- Destination: Paris
- Duration: 7 days
- Result: Complete Paris itinerary with daily activities

### Multi-City European Tour
- Source: London
- Destinations: Paris, Amsterdam, Berlin
- Duration: 14 days
- Result: Optimized route with best travel mode for each leg

### AI-Suggested Stops
1. Enter: Los Angeles → New York (10 days)
2. Click "Get AI Suggestions for Multi-City Route"
3. Claude suggests: Las Vegas, Denver, Chicago
4. Add suggested cities to customize your route

## 📁 New Files Created

### Backend
```
/trips/claude_service.go     - Claude API integration
/trips/ai_controllers.go     - API endpoints
/trips/routers.go            - Updated with new routes
```

### Frontend
```
/src/components/AITripModal.js  - Modal component
/src/pages/Dashboard.js         - Updated with AI button
/src/services/api.js            - Updated with AI endpoints
```

### Documentation
```
/AI_TRIP_PLANNER_SETUP.md      - Complete documentation
/QUICK_START_AI_FEATURE.md     - This file
```

## 🔧 Troubleshooting

### "ANTHROPIC_API_KEY not configured"
**Fix**: Add API key to `.env` file in backend directory and restart server

### Modal doesn't open
**Fix**: Check browser console for errors. Ensure frontend build succeeded.

### "Failed to generate trip"
**Check**:
1. API key is valid
2. Internet connection is working
3. Dates are in correct format (future dates)
4. Backend logs for detailed error

### Trip not saving after confirmation
**Check**:
1. Database is running
2. User is authenticated (check JWT token)
3. Backend logs for transaction errors

## 🎨 UI Components

### Main Button
- Location: Dashboard hero section
- Color: Gradient orange/coral (accent color)
- Icon: Sparkles ✨
- Text: "AI Trip Planner"

### Modal Sections
1. **Input Form**: Source, destinations, dates, preferences
2. **Preview**: Generated trip with all details
3. **Loading**: Progress indicator while Claude generates

## 📊 Example Response

When you generate a trip, Claude returns:
- Trip name & description
- Recommended travel mode (flight/train/car/bus/mixed)
- Budget breakdown by category
- 3-7 trip hops (destinations) with:
  - City details
  - Duration
  - Top attractions
  - Restaurant suggestions
- Day-by-day itinerary with:
  - 4-6 activities per day
  - Timing and duration
  - Cost estimates
  - Practical tips

## 🚀 Next Steps

After basic setup:
1. Customize Claude prompts in `claude_service.go:buildTripGenerationPrompt()`
2. Adjust UI styling in `AITripModal.js`
3. Add more preference tags
4. Integrate with real-time pricing APIs
5. Add image previews for destinations

## 💡 Pro Tips

1. **Be Specific**: More details = better itinerary
2. **Use Preferences**: Tags help Claude understand your interests
3. **Budget Matters**: Including budget gets better cost-aware suggestions
4. **Multi-City**: Use AI suggestions to discover new destinations
5. **Preview First**: Always review before confirming (you can edit later)

## 📝 Environment Variables Reference

```bash
# Required
ANTHROPIC_API_KEY=sk-ant-...

# Already configured (don't change)
DB_URL=postgres://...
SECRET=your_jwt_secret
MAPBOX_TOKEN=your_mapbox_token
```

## ⚠️ Important Notes

1. **Cost**: Each trip generation costs ~$0.02-0.05 (Haiku pricing)
2. **Rate Limits**: Check your Anthropic tier limits
3. **Security**: Never commit `.env` to git
4. **Production**: Add caching and rate limiting before deploying

## 🎉 Success Checklist

- [ ] Backend running on port 8080
- [ ] Frontend running on port 3000
- [ ] Can login to dashboard
- [ ] "AI Trip Planner" button visible
- [ ] Modal opens when clicked
- [ ] Can fill form and generate trip
- [ ] Preview shows complete itinerary
- [ ] Can save trip to database
- [ ] Trip details page shows all generated content

## 🆘 Need Help?

Check the complete documentation: `AI_TRIP_PLANNER_SETUP.md`

Common issues:
1. **API Key**: Double-check it's correctly set in `.env`
2. **Database**: Ensure PostgreSQL is running
3. **Ports**: Make sure 8080 and 3000 are available
4. **Browser**: Try Chrome/Firefox with console open for errors

---

**Ready to plan your first AI-powered trip!** 🌍✈️
